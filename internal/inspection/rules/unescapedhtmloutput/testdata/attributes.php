<?php
echo '<p title="unfinished >' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error> . '">';
echo '<p>' . '<a title="unfinished >' . <error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error> . '">';
echo "<p title=\"{<error descr="Escape untrusted text for its HTML output context.">$_GET['label']</error>}\">";
echo "<input value='" . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['label'], ENT_COMPAT)</error> . "'>";
echo '<p data-value="' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['label'], ENT_NOQUOTES)</error> . '">';
print '<p aria-label="' . <error descr="Escape untrusted text for its HTML output context.">$_POST['label']</error> . '">';
echo '<p title="' . htmlspecialchars($_GET['label'], ENT_COMPAT) . '">';
echo "<p title='" . htmlspecialchars($_GET['label'], ENT_QUOTES) . "'>";
echo "<p title='" . htmlentities($_GET['label'], ENT_QUOTES | ENT_SUBSTITUTE) . "'>";
echo '<input value="' . rawurlencode($_GET['label']) . '">';
echo '<input value="' . base64_encode($_GET['label']) . '">';
echo '<input value="' . bin2hex($_GET['label']) . '">';
echo '<input value="' . urlencode($_GET['label']) . '">';
