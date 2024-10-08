<?php
$requestPath = isset($requestPathInfo) ? $requestPathInfo : $_SERVER['REQUEST_URI'];
$urlComponents = parse_url($requestPath);
$path = $urlComponents['path'];
$scriptPath = '/xo/';

if (substr($path, 0, strlen($scriptPath)) === $scriptPath) {
    $request = substr($path, strlen($scriptPath));
    $request = $request ?: '/';
} else {
    $request = $path;
}
$isAjax = isset($_GET['ajax']);
$cssURL = "static/style.css?" . time();
$dateTime = strtolower(date('l'));
function getContent($request) {
switch ($request) {
case '/':
case '/index':
return <<<HTML
<div class="navigation"><table class="horizon"><tr><td>
<nav>✰ home / <a href="/about" onclick="loadContent('/about'); return false;">about</a> / <a href="/navi" onclick="loadContent('/navi'); return false;">navi</a> ✰</nav>
</td>
</tr>
</table>
</div>
<div class="container">
<table class="primitive">
<tr>
<td>
<p>hello, world<br>
welcome to ur homepage</p>
<p><i>what is xomud?</i></p>
<p>xomud is a multi-user hypermedia agency</p>
<button onclick="navi(alice, 'alice.rom.enclose_draggable(alice.dvr.testkit_menu_html)', 'document.body')">Click Me</button>
<p>join discord using this <a href= "https://discord.gg/9U48T5UNJN"> hyperlink</a></p>
<p><u><b>everything is subject to change</b></u></p>
</td>
</tr>
</table>
</div>
<div class="container">
<table class="primitive">
<tr>
<td>
<p>u may be here to visit our premiere vaporware ux solution.
<br> it is currently under development, <a href="/navi" onclick="loadContent('/navi'); return false;">click here to learn more.</a></p>
</td>
</tr>
</table>
</div>
HTML;
case '/about':
return <<<HTML
<div class="navigation"><table class="horizon"><tr><td>
<nav>✰ <a href="/" onclick="loadContent('/'); return false;">home</a> / about / <a href="/navi" onclick="loadContent('/navi'); return false;">navi</a> ✰</nav>
</td>
</tr>
</table>
</div>
<div class="container">
<table class="primitive">
<tr>
<td>
<p>
<img src="static/dialup.png" style="width: 28%;">
<br>
<img src="static/wthought.png" style="width: 25%;"></p>
<p><u><b>everything is subject to change</b></u></p>
</td>
</tr>
</table>
</div>
<div class="container">
<table class="primitive">
<tr>
<td>
<p>xomud.quest does not track any userdata
<br>it routes channel broadcasts and can act as a signaling server for peer-to-peer connections</p>
<p>this specific html document is generated by our web server
<br>it is routed to /xo/ as a default channel for all users</p>
<p>this document also imports our latest navi release which is given a portal address: <i>star.xomud.quest</i>
<br>this portal has a /quest api which forwards requests to a network that returns aux data for navi users</p>
<p>navi can be configured to any portal, and eventually our portal will connect to an open network
<br>for now, servers must privately register to join our portal network to handle requests made to /quest</p>
</td>
</tr>
</table>
</div>
HTML;
case '/navi':
return <<<HTML
<div class="navigation"><table class="horizon"><tr><td>
<nav>✰ <a href="/" onclick="loadContent('/'); return false;">home</a> / <a href="/about" onclick="loadContent('/about'); return false;">about</a> / navi ✰</nav>
</td>
</tr>
</table>
</div>
<div class="container">
<table class="primitive">
<tr>
<td>
<p><img src="static/meridian/01.png"><br>
<img src="static/meridian/02.png"><br>
<img src="static/meridian/03.png"><br>
<img src="static/meridian/04.png"></p>
<p><u><b>everything is subject to change</b></u></p>
</td>
</tr>
</table>
</div>
HTML;
default:
http_response_code(404);
return <<<HTML
<div class="navigation"><table class="horizon"><tr><td>
<nav>✰ <a href="/" onclick="loadContent('/'); return false;">home</a> / <a href="/about"  onclick="loadContent('/about'); return false;">about</a> / <a href="/navi" onclick="loadContent('/navi'); return false;">navi</a> ✰</nav>
</td>
</tr>
</table>
</div>
<div class="container">
<br>
<h1>404</h1>
<table class="primitive">
<tr>
<td>
<p>page not found?!?!</p>
</td>
</tr>
</table>
</div>
HTML;
}
}
$content = getContent($request);
if ($isAjax) {
echo $content;
exit;
}
?>
<!DOCTYPE html>
<html>
<head>
    <script src="https://unpkg.com/simpercell@2.0.0/navi.js"></script>
    <meta portal="https://star.xomud.quest" aux="testkit" chan="/xo/">
    <title>★ xomud.quest ★</title>
    <link rel="stylesheet" type="text/css" href="<?php echo $cssURL; ?>">
    <link rel="manifest" href="/manifest.json">
    <meta property="og:title" content="XOMUD" />
    <meta property="og:description" content="a real-time multiplayer hypermedia agency" />
    <meta property="og:image" content="https://star.xomud.quest/quest/chan/xo/static/logo.png" />
    <meta property="og:url" content="https://xomud.quest/" />
    <meta property="og:type" content="website" />
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
    <meta name="apple-mobile-web-app-capable" content="yes">
</head>
<body>
    <div class="header">
        <h1><a href="/"><img src="static/logo.png"></a></h1>
        <h1 style="color: yellow; text-shadow: 1px 1px 2px black;">xomud</h1>
    </div>
    <span id="content"><?php echo $content; ?></span>
    <div class="container">
        <table class="primitive">
            <tr>
                <td>
                    <span id="datetime">
                        <?php echo $dateTime; ?><hr>
                    </span>
                    <a href="static/compliance_certificate_xomud.png.sig.asc.html"><img src="static/compliance_certificate_xomud.png" style="width: 30%;"></a>
                </td>
            </tr>
        </table>
    </div>
    <script>
    function loadContent(url) {
        const fullUrl = window.location.origin + window.location.pathname + url;
        fetch(fullUrl + '?ajax=1')
            .then(response => response.text())
            .then(html => {
                document.getElementById('content').innerHTML = html;
                window.history.pushState({ path: fullUrl }, '', fullUrl);
            })
            .catch(error => console.error('Error loading the page: ', error));
    }
    window.onpopstate = function(event) {
        fetch(window.location.pathname + '?ajax=1')
            .then(response => response.text())
            .then(html => document.getElementById('content').innerHTML = html)
            .catch(error => console.error('Error loading the page: ', error));
    };
    </script>
    <div id="end"></div>
</body>
</html>