# macOS

## Blame

![blame/img.jpg](blame/img.jpg)

A tool computing the total memory usage per program.

### Usage

```
go run github.com/xpetit/macos/blame@latest
```

> ```
> 12484 MB         Total memory used
>
>  3006 MB   24 %  Visual Studio Code
>  2567 MB   21 %  Safari
>  1076 MB    9 %  LibreWolf
>   481 MB    4 %  Music
>   360 MB    3 %  Thunderbird
>   161 MB    1 %  WindowServer
>   137 MB    1 %  Finder
>   133 MB    1 %  CopyClip 2
> ```

## Busy

An applet showing CPU and memory usage.

100% = 1 core, 1G = 1GB

![busy/img.jpg](busy/img.jpg)

### Installation

```
git clone https://github.com/xpetit/macos.git
cd macos
go build -o busy/Busy.app/Contents/MacOS/busy ./busy
```

Open Preferences, Users & Groups, Login items, Add "Busy.app" from the folder `macos/busy`.
