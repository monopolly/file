package file

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/monopolly/dirwatch"
)

/* следит за обновлением файла */
func Notify(filename string, onUpdate func()) (err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					//fmt.Printf("File %s updated\n", filename)
					onUpdate()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println(err)
			}
		}
	}()
	err = watcher.Add(filename)
	if err != nil {
		return
	}
	<-done
	return
}

/* следит за обновлением директории */
func NotifyDir(dir string, onUpdate func(filename string)) {
	watcher := dirwatch.New(
		dirwatch.Notify(
			func(e dirwatch.Event) {
				onUpdate(e.Name)
			},
		),
	)
	defer watcher.Stop()
}
