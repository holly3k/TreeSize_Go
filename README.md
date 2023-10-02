# TreeSize_Go

TreeSize is a application used to scan a folder and display a treeview by the size of folder or files
- [x] written in golang
- [x] base on [lxn/walk](https://github.com/lxn/walk)
- [x] only work on windows

## How to compile
```
go mod tidy
go build -ldflags="-H windowsgui"
```

