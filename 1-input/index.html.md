<%! import pyhpy %>
<%block name="PAGE_CSS">
    <link rel="stylesheet" type="text/css" href="${pyhpy.url('/static/css/home.css')}">
</%block>
PyHPy is a static website generation framework, built on top of the [Python](http://python.org/)-powered [Mako](http://www.makotemplates.org/) templating engine.

TODO:

* Re-do this home-page.  It's terrible.

* "Show this page source"

* Copy Clay style -- beautiful, and home page explains everything perfectly.

* I like Nikola's docs -- showing cli screenshots to the right.  Really helps to reduce size of docs.

* Sphinx doc format is beautiful.

* Photo album should show PyHPy-related photos, helpful for better understanding.




=== <i class="fa fa-files-o fa-lg"></i> Static is Beautiful ===

Static webpages have some very nice features:

* Static is *Fast*.  Nothing is faster than serving a static page.  A static website can easily survive huge traffic spikes.

* Static is Immortal.  Once a static webpage is online, it almost never breaks.  No maintenence is required to keep them online.  They handle server upgrades, library updates, and host migrations with ease.

* Impossible to Hack.  A static page never introduces any extra security vulnerabilities.  Much, *much* safer than a dynamic site.

* Easy to Host.  Static websites are supported by *every* web host.


=== <i class="fa fa-thumbs-o-up fa-lg"></i> Mako is Awesome ===

I chose to build this framework on top of [Mako](http://www.makotemplates.org/) for the following reasons:

* Mako can produce any kind of output -- not just webpages.  Therefore, you can use PyHPy for any kind of static output generation.  (I like general solutions.)

* Mako's [powerful template inheritence](http://docs.makotemplates.org/en/latest/inheritance.html) makes it easy to create consistent, scalable websites, no matter how large your site gets.

* Small Feature Set.  Helps you to learn it quickly.

* Widely Used.  Two famous users are [reddit.com](https://github.com/reddit/reddit/) and [Pylons/Pyramid](http://www.pylonsproject.org/).


=== <i class="fa fa-paper-plane-o fa-lg"></i> PyHPy Makes it Easy ===

PyHPy combines powerful tools like `rsync`, `make`, and `mako-render`, and adds a layer of convenience and integration on top.

* [MarkDown](https://en.wikipedia.org/wiki/Markdown) Integration.  Just type your content text into an `.md` file and it becomes a beautiful webpage.  Easy enough for my Mom.

* Improved Error Reporting.  When things go wrong, PyHPy produces very detailed and helpful error reports, allowing you to pin-point the cause of the problem with minimal effort.

* TODO: I need to write more details that are specific to PyHPy.  Get rid of the Expires Header and FontAwesome stuff -- anybody can do that.  I need to answer the question: does PyHPy actually solve a real problem?  Why would I use it over Blogofile?  ...Or Pelican?  Or Jekyll?  Can I integrate with GitHub Pages?  Do I plan to support ReStructuredText?  Do I have easy ability to use themes from elsewhere?  No.  ...I guess PyHPy is supposed to be a good *general* solution to static output generation, and therefore doesn't have the ability to focus on any of those particular goals.  Need to demonstrate awesome, *realistic* use cases that the other systems just can't do.  Also seems like PyHPy actually has a much simpler model than Blogofile/Nikola.  Much less structured --> less to learn to get started, more work later.  I need to adjust PyHPy so that it's *so* simple that it can be understood *completely* in several bullet-points.

Here's the typical process of building a site with PyHPy:

1. Create content (like MarkDown), Mako templates, and static files (like images) in an `input` directory.
1. Run `make`.  Content and templates are rendered, and the results are published to an `output` directory.
1. View the output in a web browser.  If you're not totally happy with the result, GOTO 1.
1. Upload the `output` directory to your web server using `rsync` or `FTP`.

The best way to start learning PyHPy is to read through the included `input.example/`, which produces *this* website.  You can also read the documentation articles that are included in the [demo blog](${pyhpy.url('/blog.html', mtime=None)}).  If you have questions, send them to [Christopher Sebastian](mailto:csebastian3@gmail.com).

**[<i class="fa fa-photo fa-lg"></i> View the Demo Photo Album](${pyhpy.url('/photos.html', mtime=None)})**

**[<i class="fa fa-newspaper-o fa-lg"></i> View the Demo Blog](${pyhpy.url('/blog.html', mtime=None)})**

