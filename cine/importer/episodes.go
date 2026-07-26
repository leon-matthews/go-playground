package importer

import (
	"context"
	"database/sql"
	"io"

	"local.dev/cine/reader"
)

// episodeColumns are the episodes columns in the order bindEpisodeRow writes them.
var episodeColumns = []string{"id", "parent_id", "season_number", "episode_number"}

// importEpisodes streams title.episode into the episodes table, returning the
// number of episodes written.
func importEpisodes(ctx context.Context, tx *sql.Tx, episodes io.Reader) (counts, error) {
	inserter, err := newBatchInserter(ctx, tx, "episodes", episodeColumns, bindEpisodeRow)
	if err != nil {
		return counts{}, err
	}
	var read int64
	for record, err := range reader.ReadTitleEpisode(episodes) {
		if err != nil {
			return counts{}, err
		}
		read++
		row, err := buildEpisodeRow(record)
		if err != nil {
			return counts{}, err
		}
		if err := inserter.Add(ctx, row); err != nil {
			return counts{}, err
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return counts{}, err
	}
	return counts{read: read, written: inserter.Added()}, nil
}

// episodeRow holds one episodes row's values in column order; a nil field is
// stored as SQL NULL.
type episodeRow struct {
	id       int64
	parentID int64
	season   any
	episode  any
}

// buildEpisodeRow transforms a reader record into an episodes row.
func buildEpisodeRow(e reader.TitleEpisode) (episodeRow, error) {
	id, err := parseID(e.Tconst)
	if err != nil {
		return episodeRow{}, err
	}
	parentID, err := parseID(e.ParentTconst)
	if err != nil {
		return episodeRow{}, err
	}
	return episodeRow{
		id:       id,
		parentID: parentID,
		season:   nullableInt(e.SeasonNumber),
		episode:  nullableInt(e.EpisodeNumber),
	}, nil
}

// bindEpisodeRow appends an episodes row's values in episodeColumns order.
func bindEpisodeRow(args []any, r episodeRow) []any {
	return append(args, r.id, r.parentID, r.season, r.episode)
}
