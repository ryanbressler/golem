# Introduction #
![http://golem.googlecode.com/files/GolemScreenShot2.png](http://golem.googlecode.com/files/GolemScreenShot2.png)
The **golem** project provides hooks for user interfaces.  The `RestOnJob` interface provides an `/html` URL mapping that serves HTML content from the filesystem.  In addition, we are including example HTML content under http://golem.googlecode.com/hg/src/html to demonstrate how to construct a web user interface.

The example code is based on the [ExtJS](http://sencha.com/extjs) Javascript framework, but this is not a required dependency.  The javascript is loaded from [ISB](http://systemsbiology.org) servers and the [Addama project](http://addama.googlecode.com/svn/branches/2.1/user-interfaces).

# Details #
To serve HTML content from **golem** services (e.g. Master, Scribe), there must be an `html` directory in the path of the service.  For instance:
```
    cd /local/projects/golem/src
    gomake
    cp golem /local/webapps/mygolem/
    cd /local/apps/webapps/mygolem/
    mkdir html
    cp $CONTENT_HOME/* html/
    ./golem -m -hostname myhost:8083
```

The cluster ui will be available at `https://myhost:8083/html/`. Visualizations of cluster performance will be available at `https://myhost:8083//html/stream.html`.

This mechanism allows any content contained in the html directory to be served by the **golem** services.