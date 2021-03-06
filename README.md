# wiki
A personal wiki written in Go

The motivation for this project was twofold.  One I wanted a bit of a side project to improve my Golang and two I wanted a note taking app to replace Evernote.

I use markdown for the notes themselves.  They are stored more or less as is on the file system so that you can edit directly if need be (for example, I use a Dropbox folder for my wiki files that I can then access on my mobile and edit directly).

The app is very much a work in progress.  It works enough for me to use it but it but there are a lot of rough edges.  Furthermore, security is very lax so it is not something you should put on the internet.

# Building
To get up and running clone as per usual, go get to pull down any dependencies and then go build or install.

You will also need to pull down latest npm dependencies for the web stuff to work.  Cd in to static/js and npm install in there to get the latest stuff.

# Running
To be honest the config side of things needs a little work to make it more user friendly to run up for the first time - like I said, still a work in progress.

You need a valid config file in order to run the app.  Hard coded to config.json at the moment.

The config file is standard json.  Keys are as follows

|JSON KEY|ENV|DEFAULT|NOTES|
|-------|----|------|-------|
|HTTPPort | PORT | 80 | Port for non-secure pages |
|WikiDir|WIKIDIR|"wikidir"|Folder to place markdown files in - I point this at my Dropbox sync'd folders|
|Logfile|LOGFILE|"wiki.log"|File to save logging to|
|EncryptionKey|ENCRYPTIONKEY||32 character key used to encrpt files - see below|


# Getting Started

The home page allows you to create new wiki pages.  Just bung in a wiki page name and hit create.

That will take you to the edit page for that wiki word.  Type in some text - there is a handy markdown prompt sheet on the righ.

You can add tags to the page - comma separated.  Tags show up at the bottom of the menu on the left.

Before saving you can opt to encrypt and/or publish the page.  Encrypting a page will save the page as an encrypted file preventing others from reading the file on the OS.  Otherwise wiki pages are saved as plain markdown.

Publishing a page makes it available on a different URL - more on this later.

Saving the page adds it to the menu on the left.  Selecting a page on the left shows the rendered markdown version which you can then edit again.

If you use a / in your wiki page title it will appear in a folder - i.e. the folder is added to your menu and the page is listed below the folder when you select.  The menu only supports one deep at this point in time.

When editing a page you can use {{wikilink}} this will render the brackets as a link to a page on the wiki.  If the page does not exist then when you click on it you will get the edit screen for that page.  You can use / in these links as well to link to pages in folders.

You can also use a # to point to a page heading within the page you are linking to.  So, {{test/some page#a heading}} will give you a link to "a heading" on "some page" in the "test" folder.

Back on the home page you will notice a link to "Public Pages".  Any page that you mark as published will appear here.  I use this at work when I want to share a readonly copy of a page with someone - they are able to access public pages without signing in.  Also, public pages do not include the menu.  If you have included wiki links to pages that are not public these will fail

PDF files can be added to the wiki folder and they are automatically picked up and added to the menu and tagged with PDF.

Other file types can also be added although they will not get added to the menu you can still link to them.  For example, I often save PNG files in the wiki folders and then render them on wiki pages by including as html on the page using the usual img tags.  Any file can be referenced directly by using <host>:<port>/wiki/raw/image.png

Finally, on the home page there is also a rudimentary search that will search all files in all folders for your search term - think of it as a simple grep for a string.  No fancy regex support just yet.

# References

Markdown rendering courtesy of https://github.com/golang-commonmark/markdown

URL Excaping lib github.com/opennota/urlesc

# Changelog (ish)

* 2020-06: Experiments with AWS deployment.  Abandoned due to hassles of setting up EFS with Fargate - life is too short.  Went to Azure instead.
* 2020-07: I have removed all HTTPS and auth support from the app having replaced with external apps to perform authorisation.

