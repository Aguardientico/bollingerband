bollingerband
=============

How to use:
-----------
> Please be sure to have installed **git** and **mercurial** since some dependencies should be cloned from different repositories.  
> Then in a command line execute:  
`go get github.com/Aguardientico/bollingerband`

**Note**
> One of the dependencies (**draw2d**) have an issue. I've set a patch (https://code.google.com/p/draw2d/issues/detail?id=29) to fix it but still today (09/02/2014) the issue is not fixed
If you got the following message: `go/src/code.google.com/p/draw2d/draw2d/image.go:166: cannot use nil as type truetype.Hinting in function argument`
then you should apply manually the patch and run:  
`go install github.com/Aguardientico/bollingerband`

then you can just run
`bollingerband`
It should generate one pgn file for each symbol

if you want to change symbols, periods or factor for bollinger band then you can change it in *config.json* file
