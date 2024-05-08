# Sony media organizer tool

Install:

```
go install github.com/fgazat/sony
```


Usage:

```
sony sort -dst="DST_DIR" [PATH_TO_FOLDER_WITH_MEDIA]
```

It will group standart Sony file structure to: 

```
DST_DIR/
├── ARW/
├── JPEG/
└── MP4/
```

Original files from `PATH_TO_FOLDER_WITH_MEDIA` won't be edited or deleted.

