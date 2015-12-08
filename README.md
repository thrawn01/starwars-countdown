# Wut iz this?

This is the code that runs http://starwars-countdown.com

# How do I build it?
```
$ cd $GOPATH/src
$ git clone github.com/thrawn01/starwars-countdown.git
$ cd starwars-countdown
$ go build
```

# How do I run it?
```
./starwars-countdown -h
Usage:
  starwars-countdown [OPTIONS]

Application Options:
  -b, --bind=
  -i, --image-dir=    Location of the images within the public-dir (images/) [$SWCD_IMAGE_DIR]
  -p, --public-dir=   The directory where index.html lives (public/) [$SWCD_PUBLIC_DIR]
  -o, --output-index  Print index.html to stdout and exit
  -d, --debug

Help Options:
  -h, --help          Show this help message
```

The server will look in ```public/images``` for images to put into the slideshow.
You have to supply the images locally, however if you don't want to serve
images locally you can specify an image url by creating a one line file with a
url in it. 

Like so
```
$ echo "http://i.imgur.com/YzNPZH8.jpg" > x-wing.lnk
```

Now when the server starts it will put the [http://i.imgur.com/YzNPZH8.jpg](http://i.imgur.com/YzNPZH8.jpg) into the slide show.
