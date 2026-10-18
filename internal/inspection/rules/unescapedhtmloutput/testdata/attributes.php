<?php
<error descr="Escape untrusted text for its HTML output context.">echo '<p title="unfinished >' . $_GET['label'] . '">';</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . '<a title="unfinished >' . $_GET['label'] . '">';</error>
<error descr="Escape untrusted text for its HTML output context.">echo "<p title=\"{$_GET['label']}\">";</error>
<error descr="Escape untrusted text for its HTML output context.">echo "<input value='" . htmlspecialchars($_GET['label'], ENT_COMPAT) . "'>";</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<p data-value="' . htmlspecialchars($_GET['label'], ENT_NOQUOTES) . '">';</error>
<error descr="Escape untrusted text for its HTML output context.">print '<p aria-label="' . $_POST['label'] . '">'</error>;
echo '<p title="' . htmlspecialchars($_GET['label'], ENT_COMPAT) . '">';
echo "<p title='" . htmlspecialchars($_GET['label'], ENT_QUOTES) . "'>";
echo "<p title='" . htmlentities($_GET['label'], ENT_QUOTES | ENT_SUBSTITUTE) . "'>";
echo '<input value="' . rawurlencode($_GET['label']) . '">';
echo '<input value="' . base64_encode($_GET['label']) . '">';
echo '<input value="' . bin2hex($_GET['label']) . '">';
echo '<input value="' . urlencode($_GET['label']) . '">';
