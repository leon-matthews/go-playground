// CLI interface to blogposts functionality
package main

import (
	"log"
	"os"

	blogposts "reading_files"
)

func main() {
	posts, err := blogposts.NewPostsFromFS(os.DirFS("posts"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(posts)
}
