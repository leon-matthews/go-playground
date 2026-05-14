// Command audiobook-chapterise breaks a single-file audio book into separate
// MP3 files.
//
// Mirrors audiobook-chapterise.py: read chapter metadata via ffprobe (falling
// back to evenly-sized parts), preview the planned output and ask for
// confirmation, then extract each clip with ffmpeg in parallel.
//
// Requires `ffmpeg` and `ffprobe` on $PATH. Build with `go build` or run
// directly with `go run audiobook-chapterise.go PATH`.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/spf13/pflag"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const (
	targetMinutes = 20
	mp3Quality    = 6
)

// ---- filenames ----

var (
	illegalCharRe = regexp.MustCompile(`[^\p{L}\p{N}_' .,()]`)
	whitespaceRe  = regexp.MustCompile(`\s+`)
)

func cleanFilename(name string) string {
	s := strings.TrimSpace(name)
	s = strings.ReplaceAll(s, ":", " - ")
	s = illegalCharRe.ReplaceAllString(s, " ")
	s = whitespaceRe.ReplaceAllString(s, " ")
	return s
}

// ---- seconds ----

type seconds float64

func (s seconds) split() (int, int, float64) {
	total := float64(s)
	h := int(math.Floor(total / 3600))
	rem := total - float64(h)*3600
	m := int(math.Floor(rem / 60))
	return h, m, rem - float64(m)*60
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

func (s seconds) humanDuration() string {
	h, m, _ := s.split()
	if h > 0 {
		return fmt.Sprintf("%s and %s", plural(h, "hour"), plural(m, "minute"))
	}
	return plural(m, "minute")
}

// ---- ffprobe / ffmpeg ----

type ffprobeChapter struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Tags      struct {
		Title string `json:"title"`
	} `json:"tags"`
}

type ffprobeData struct {
	Chapters []ffprobeChapter `json:"chapters"`
	Format   struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func ffprobe(ctx context.Context, path string, verbose bool) (*ffprobeData, error) {
	args := []string{
		"ffprobe",
		"-hide_banner",
		"-loglevel", "warning",
		"-i", path,
		"-print_format", "json",
		"-show_chapters",
		"-show_format",
	}
	out, err := runCmd(ctx, args, verbose)
	if err != nil {
		return nil, fmt.Errorf("ffprobe: %w", err)
	}
	var data ffprobeData
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("invalid ffprobe JSON: %w", err)
	}
	return &data, nil
}

func ffmpegExtractAudio(ctx context.Context, mediaPath string, start, end float64, output string, verbose bool, onProgress func(microseconds int64)) error {
	args := []string{
		"ffmpeg",
		"-nostdin",
		"-hide_banner",
		"-loglevel", "error",
		"-progress", "pipe:1",
		"-i", mediaPath,
		"-vn", "-sn", "-dn",
		"-ss", fmt.Sprintf("%.3f", start),
		"-to", fmt.Sprintf("%.3f", end),
		"-codec:a", "libmp3lame",
		"-ac", "2",
		"-qscale:a", strconv.Itoa(mp3Quality),
		"-f", "mp3",
		"-n",
		output,
	}
	if verbose {
		fmt.Fprintln(os.Stderr, shellJoin(args))
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "Command %q not found on system. Please install.\n", args[0])
			os.Exit(100)
		}
		return err
	}

	progDone := make(chan struct{})
	go func() {
		defer close(progDone)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			k, v, ok := strings.Cut(scanner.Text(), "=")
			if !ok || k != "out_time_us" {
				continue
			}
			us, parseErr := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if parseErr != nil || onProgress == nil {
				continue
			}
			onProgress(us)
		}
	}()

	waitErr := cmd.Wait()
	<-progDone

	if waitErr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			msg := strings.TrimSpace(stderr.String())
			return fmt.Errorf("command returned error code %d: %q", exitErr.ExitCode(), msg)
		}
		return waitErr
	}
	return nil
}

// ---- chapters ----

type Chapter struct {
	Start float64
	End   float64
	Title string
}

type mediaInfo struct {
	path string
	data *ffprobeData
}

func newMediaInfo(ctx context.Context, path string, verbose bool) (*mediaInfo, error) {
	data, err := ffprobe(ctx, path, verbose)
	if err != nil {
		return nil, err
	}
	return &mediaInfo{path: path, data: data}, nil
}

func (m *mediaInfo) duration() (float64, error) {
	return strconv.ParseFloat(m.data.Format.Duration, 64)
}

func (m *mediaInfo) chapters() []Chapter {
	out := make([]Chapter, 0, len(m.data.Chapters))
	for _, c := range m.data.Chapters {
		start, _ := strconv.ParseFloat(c.StartTime, 64)
		end, _ := strconv.ParseFloat(c.EndTime, 64)
		out = append(out, Chapter{Start: start, End: end, Title: c.Tags.Title})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start < out[j].Start })
	return out
}

type chapterPlanner struct {
	media    *mediaInfo
	duration float64
	start    int
}

func newChapterPlanner(ctx context.Context, path string, start int, verbose bool) (*chapterPlanner, error) {
	mi, err := newMediaInfo(ctx, path, verbose)
	if err != nil {
		return nil, err
	}
	d, err := mi.duration()
	if err != nil {
		return nil, fmt.Errorf("could not read media duration: %w", err)
	}
	return &chapterPlanner{media: mi, duration: d, start: start}, nil
}

func (p *chapterPlanner) plan() []Chapter {
	chs := p.media.chapters()
	if len(chs) == 0 {
		fmt.Fprintf(os.Stderr, "No chapters found in audio file: %q\n", filepath.Base(p.media.path))
		chs = p.makeParts()
	}
	return chs
}

func (p *chapterPlanner) makeParts() []Chapter {
	numParts := max(int(math.Round((p.duration/60)/float64(targetMinutes))), 1)
	step := p.duration / float64(numParts)
	parts := make([]Chapter, 0, numParts)
	start := 0.0
	end := step
	for i := range numParts {
		parts = append(parts, Chapter{
			Start: start,
			End:   end,
			Title: fmt.Sprintf("Part %d", i+p.start),
		})
		start = end
		end += step
	}
	return parts
}

// ---- naming ----

type clipNamer struct {
	start   int
	padding int
}

func newClipNamer(start, numChapters int) *clipNamer {
	return &clipNamer{
		start:   start,
		padding: calculatePadding((numChapters - 1) + start),
	}
}

func calculatePadding(maxValue int) int {
	p := max(int(math.Ceil(math.Log10(float64(maxValue+1)))), 2)
	return p
}

func (n *clipNamer) filename(index int, ch Chapter, suffix string) string {
	prefix := index + n.start
	raw := fmt.Sprintf("%0*d. %s.%s", n.padding, prefix, strings.TrimSpace(ch.Title), suffix)
	return cleanFilename(raw)
}

func (n *clipNamer) foldername(stem string) string {
	return cleanFilename(stem)
}

// ---- writing ----

type clipWriter struct {
	mediaPath string
	folder    string
}

func (w *clipWriter) createFolder() error {
	if _, err := os.Stat(w.folder); err == nil {
		return fmt.Errorf("output folder already exists: %q", w.folder)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Mkdir(w.folder, 0o755)
}

func (w *clipWriter) write(ctx context.Context, ch Chapter, outputPath string, verbose bool, onProgress func(microseconds int64)) error {
	partial := outputPath + ".part"
	if err := ffmpegExtractAudio(ctx, w.mediaPath, ch.Start, ch.End, partial, verbose, onProgress); err != nil {
		return err
	}
	return os.Rename(partial, outputPath)
}

// ---- shell / process helpers ----

func runCmd(ctx context.Context, args []string, verbose bool) ([]byte, error) {
	if verbose {
		fmt.Fprintln(os.Stderr, shellJoin(args))
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		var execErr *exec.Error
		if errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "Command %q not found on system. Please install.\n", args[0])
			os.Exit(100)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			msg := strings.TrimSpace(stderr.String())
			return nil, fmt.Errorf("command returned error code %d: %q", exitErr.ExitCode(), msg)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

const shellSafeChars = "@%+=:,./-_"

func shellJoin(args []string) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = shellQuote(a)
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	for _, r := range s {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune(shellSafeChars, r)) {
			return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
		}
	}
	return s
}

// ---- preview ----

func preview(chapters []Chapter, namer *clipNamer, folder string, duration float64) bool {
	files := make([]string, len(chapters))
	for i, c := range chapters {
		files[i] = namer.filename(i, c, "mp3")
	}
	total := seconds(duration)
	avg := seconds(duration / float64(len(files)))
	fmt.Printf("Creating %d files averaging %s each,\n", len(files), avg.humanDuration())
	fmt.Printf("a total of %s of audio:\n", total.humanDuration())
	fmt.Printf("📁 %s/\n", filepath.Base(folder))

	printColumns(os.Stdout, files, terminalWidth())
	fmt.Println()

	return confirmYes("Do you wish to proceed?", true)
}

func printColumns(out *os.File, items []string, width int) {
	if len(items) == 0 {
		return
	}
	maxLen := 0
	for _, s := range items {
		if l := utf8.RuneCountInString(s); l > maxLen {
			maxLen = l
		}
	}
	colWidth := maxLen + 2
	cols := max(width/colWidth, 1)
	rows := (len(items) + cols - 1) / cols
	for r := range rows {
		for c := range cols {
			idx := c*rows + r
			if idx >= len(items) {
				break
			}
			s := items[idx]
			fmt.Fprint(out, s)
			if c < cols-1 {
				pad := colWidth - utf8.RuneCountInString(s)
				fmt.Fprint(out, strings.Repeat(" ", pad))
			}
		}
		fmt.Fprintln(out)
	}
}

func confirmYes(prompt string, defaultYes bool) bool {
	yn := "[Y/n]"
	if !defaultYes {
		yn = "[y/N]"
	}
	fmt.Printf("%s %s ", prompt, yn)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}
	line = strings.TrimSpace(strings.ToLower(line))
	switch line {
	case "":
		return defaultYes
	case "y", "yes":
		return true
	default:
		return false
	}
}

// ---- terminal ----

type winsize struct {
	Row, Col, Xpixel, Ypixel uint16
}

func terminalWidth() int {
	for _, fd := range []uintptr{uintptr(syscall.Stderr), uintptr(syscall.Stdout), uintptr(syscall.Stdin)} {
		ws := winsize{}
		_, _, errno := syscall.Syscall(
			syscall.SYS_IOCTL,
			fd,
			uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(&ws)),
		)
		if errno == 0 && ws.Col > 0 {
			return int(ws.Col)
		}
	}
	if c := os.Getenv("COLUMNS"); c != "" {
		if w, err := strconv.Atoi(c); err == nil && w > 0 {
			return w
		}
	}
	return 80
}

// ---- parallel encode ----

const labelWidth = 32

func runEncode(ctx context.Context, writer *clipWriter, namer *clipNamer, chapters []Chapter, jobs int, verbose bool) error {
	jobWord := "jobs"
	if jobs == 1 {
		jobWord = "job"
	}

	p := mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithWidth(40),
	)

	totalBar := p.New(int64(len(chapters)),
		mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding(" ").Rbound("]"),
		mpb.BarPriority(math.MaxInt),
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("Encoding (%d %s)", jobs, jobWord), decor.WC{W: labelWidth}),
			decor.CountersNoUnit("%d/%d"),
		),
		mpb.AppendDecorators(
			decor.Elapsed(decor.ET_STYLE_HHMMSS, decor.WC{W: 10}),
		),
	)

	sem := make(chan struct{}, jobs)
	var (
		wg       sync.WaitGroup
		errMu    sync.Mutex
		firstErr error
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	setErr := func(err error) {
		errMu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMu.Unlock()
	}

loop:
	for i, ch := range chapters {
		select {
		case <-ctx.Done():
			break loop
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func(i int, ch Chapter) {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				return
			}

			durUs := int64((ch.End - ch.Start) * 1_000_000)
			label := truncateLabel(ch.Title, labelWidth-1)
			bar := p.New(durUs,
				mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding(" ").Rbound("]"),
				mpb.BarRemoveOnComplete(),
				mpb.PrependDecorators(
					decor.Name(label, decor.WC{W: labelWidth}),
				),
				mpb.AppendDecorators(
					decor.Percentage(decor.WC{W: 5}),
				),
			)

			output := filepath.Join(writer.folder, namer.filename(i, ch, "mp3"))
			err := writer.write(ctx, ch, output, verbose, func(us int64) {
				bar.SetCurrent(us)
			})
			if err != nil {
				bar.Abort(true)
				if !errors.Is(err, context.Canceled) {
					setErr(err)
				}
				cancel()
				return
			}
			bar.SetCurrent(durUs)
			totalBar.Increment()
		}(i, ch)
	}
	wg.Wait()
	p.Wait()
	return firstErr
}

func truncateLabel(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n-1]) + "…"
}

// ---- options ----

type options struct {
	jobs    int
	start   int
	verbose bool
	confirm bool
	path    string
}

func parseFlags(argv []string) (*options, error) {
	defaultJobs := max(runtime.NumCPU()/2, 1)

	fs := pflag.NewFlagSet("audiobook-splitter", pflag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] PATH\n\n", fs.Name())
		fmt.Fprintln(os.Stderr, "Break audio book into multiple files, one per chapter.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	o := &options{}
	fs.IntVarP(&o.jobs, "jobs", "j", defaultJobs, "number of ffmpeg jobs to run in parallel")
	fs.IntVarP(&o.start, "start", "s", 1, "starting chapter number")
	fs.BoolVarP(&o.verbose, "verbose", "v", false, "print commands as they are run")
	var skipConfirm bool
	fs.BoolVarP(&skipConfirm, "yes", "y", false, "assume yes; do not ask for confirmation")
	if err := fs.Parse(argv); err != nil {
		return nil, err
	}
	o.confirm = !skipConfirm
	if o.jobs < 1 {
		return nil, errors.New("--jobs must be at least 1")
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return nil, errors.New("expected exactly one PATH argument")
	}
	o.path = fs.Arg(0)
	return o, nil
}

// ---- main ----

func main() {
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(runMain(ctx, opts))
}

func runMain(ctx context.Context, opts *options) int {
	absPath, err := filepath.Abs(opts.path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	planner, err := newChapterPlanner(ctx, absPath, opts.start, opts.verbose)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "\nInterrupted.")
			return 130
		}
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	chapters := planner.plan()
	if len(chapters) == 0 {
		fmt.Fprintln(os.Stderr, "No chapters to extract.")
		return 1
	}
	namer := newClipNamer(opts.start, len(chapters))

	base := filepath.Base(absPath)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	folder := filepath.Join(filepath.Dir(absPath), namer.foldername(stem))
	writer := &clipWriter{mediaPath: absPath, folder: folder}

	if opts.confirm {
		if !preview(chapters, namer, folder, planner.duration) {
			return 0
		}
	}

	if err := writer.createFolder(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	jobs := min(opts.jobs, len(chapters))
	if err := runEncode(ctx, writer, namer, chapters, jobs, opts.verbose); err != nil {
		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "\nInterrupted.")
			return 130
		}
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
