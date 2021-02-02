package fixtures

// HTMLWithBadLoader is a sample of a html response that does not contain the Browser loader script in the right place(outside of </head> tag)
var HTMLWithBadLoader = `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML+RDFa 1.1//EN">
<html lang="en" dir="ltr" version="HTML+RDFa 1.1"
  xmlns:xsd="http://www.w3.org/2001/XMLSchema#">
<head profile="http://www.w3.org/1999/xhtml/vocab">
<!-- NR header -->
<div class="container-fluid">
<title>Documentation | New Relic Documentation</title>
<meta charset="utf-8" />
<meta name="description" content="New Relic Documentation: Table of Contents" />
<script src="https://www.google.com/recaptcha/api.js?hl=en" async="async" defer="defer"></script>
<link rel="alternate" href="https://docs.newrelic.co.jp/" hreflang="ja" />
<meta class="swiftype" name="translation_ja_url" data-type="string" content="https://docs.newrelic.co.jp/" />  <meta name="google-site-verification" content="Gz159HQpDMScxkpvmeDAU5tSbRpLHW0c1iimYcDkNR4" />
<meta name="viewport" content="width=device-width, height=device-height, initial-scale=1.0, user-scalable=0, minimum-scale=1.0, maximum-scale=1.0">
<link type="text/css" rel="stylesheet" href="https://docs.newrelic.com/sites/default/files/css/css_xE-rWrJf-fncB6ztZfd2huxqgxu4WO-qwma6Xer30m4.css" media="all" />
</head>
<body ontouchstart="" class="html front not-logged-in one-sidebar sidebar-first page-node page-node- page-node-446 node-type-page i18n-en"  >
<script type="text/javascript">(window.NREUM||(NREUM={})).loader_config={xpid:"UwQAUVNaGwcGVFVbBwg=",licenseKey:"29bb765936",applicationID:"4604909"};window.NREUM||(NREUM={}),_</script>  
<div id="page-wrapper">
  <div id="page-home">
    <div id="header-stub"></div>

<header class="banner" id="header-home" role="banner">

  <div class="row no-margin">
    <div class="brand-box col-md-2">
      <button type="button" class="navbar-toggle pull-right collapsed" data-toggle="collapse" data-target="#header-main-menu" aria-expanded="false">
        <i class="fa fa-bars fa-2x" aria-hidden="true"></i>
      </button>
      <button type="button" class="navbar-toggle search-toggle pull-right collapsed" data-toggle="collapse" data-target="#search-box" aria-expanded="false">
        <i class="fa fa-search fa-2x" aria-hidden="true"></i>
      </button>
      <div id="logo" class="">
        <a class="header-logo-img" href="/" title="Home" rel="home" id="">
            <img class="img-responsive" src="https://docs.newrelic.com/sites/default/files/thumbnails/image/Main%20logo.png" alt="Home">
        </a>
      </div>
    </div>
</header> 
<script type="text/javascript" src="https://docs.newrelic.com/sites/default/files/js/js_W4vgHF9MX0jnCNVma5w6hmmbGoToFf6nuWAyClDOFYo.js"></script>
<script src="/sites/all/libraries/newrelic/swiftype_lib.js"></script>
<script type="text/javascript">window.NREUM||(NREUM={});NREUM.info={"beacon":"bam.nr-data.net","licenseKey":"29bb765936","applicationID":"4604909","transactionName":"YQMHY0NRCEpTAhUPWlhJJFRFWQlXHQ8OAlBpFgRQVG8QUFcW","queueTime":0,"applicationTime":159,"atts":"TUQQFQtLREtCDD4SWl0DC2hFSRZcEFtDB1ZVAxZEExxES0IMPhJaXQMLaEdVFEpbDg9EDxRXRxsTQhZUbQgSOVtTERdSXVkFZlMFDA9bFFxHBhMcREtCDD4DWFcPCRULEgRSRwIEH3VYAxJFVFwPWhwCDgsXGkQXR1xvE0pXEz4PURRcRwYJA14BCldDShdDFQBFblEBXFwVQ1wXewkfXl1cB2UdVE9WFR4rBFRYXhJWQQlaRnxYEgBbEX0HWhIuMkZtFldVaAAEOQ8bQSAWRVoDMlJTew9Nbk5UVQIYVVMXGXsubX8tTUZZXw0AF3ZVBVJdSEElXUQJCFJtH14IHFFPUgUCUksGAANGalMHABRcaklQBAYeVQ8QTUMOWkUSC1ZcVUQDEFBWVBsHUEsHHwJUGx5DEQdSUzkQRV0SXBtaFRUWRgw6SmseVAlaQU8PA0JEAwleUh4FVl89TkQZFBQAUVRCA0ttFBMKFwxER0odEgcbCBocGw==","errorBeacon":"bam.nr-data.net","agent":""}</script>
</body>
</html>`
