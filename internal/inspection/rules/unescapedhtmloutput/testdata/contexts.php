<?php
// Contexts where escaping for HTML text is not enough.
echo '<script>' . ('<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>);
echo '<sty' . 'le>' . ('<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>);
echo "<p\x3e\x3cscript>{<error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>}";
echo '<script>let label = ' . json_encode($_GET['label']) . ';</script>';
echo '<script>let label = ' . <error descr="Escape untrusted text for its HTML output context.">json_encode($_GET['label'], JSON_UNESCAPED_SLASHES)</error> . ';</script>';
echo '<script>let label = ' . json_encode($_GET['label'], JSON_UNESCAPED_SLASHES | JSON_HEX_TAG) . ';</script>';
echo '<script>let page = ' . (int) $_GET['page'] . ';</script>';
echo '<a href="' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['url'], ENT_QUOTES)</error> . '">link</a>';
echo '<a href="/search?q=' . htmlspecialchars($_GET['q'], ENT_QUOTES) . '">link</a>';
echo '<a href="' . rawurlencode($_GET['q']) . '">link</a>';
echo '<button onclick="go(\'' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['q'], ENT_QUOTES)</error> . '\')">go</button>';
echo '<button onclick="go(' . htmlspecialchars(json_encode($_GET['q']), ENT_QUOTES) . ')">go</button>';
echo '<p style="color: ' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['color'], ENT_QUOTES)</error> . '">text</p>';
echo '<iframe srcdoc="' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['html'], ENT_QUOTES)</error> . '"></iframe>';
echo '<p class=' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['class'], ENT_QUOTES)</error> . '>text</p>';
echo '<p class=' . rawurlencode($_GET['class']) . '>text</p>';
echo '<p data-id=' . (int) $_GET['id'] . '>text</p>';
echo '<a href="javascript:' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['code'], ENT_QUOTES)</error> . '">run</a>';
echo '<a href="java&#9;script:' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['code'], ENT_QUOTES)</error> . '">run</a>';
echo '<a href="https://example.com/?q=' . htmlspecialchars($_GET['q'], ENT_QUOTES) . '">link</a>';
echo '<script>let label = "' . <error descr="Escape untrusted text for its HTML output context.">json_encode($_GET['label'])</error> . '";</script>';
echo '<script>/* label */ let label = ' . json_encode($_GET['label']) . ';</script>';
echo '<button onclick="go(&quot;' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars(json_encode($_GET['q']), ENT_QUOTES)</error> . '&quot;)">go</button>';
