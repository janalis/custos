<?php
// Every request-derived part of an output statement is reported on its own.
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['first']</error> . '</p><p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['second']</error> . '</p>';
echo '<p>', <error descr="Escape untrusted text for its HTML output context.">$_GET['first']</error>, '</p><a title="', <error descr="Escape untrusted text for its HTML output context.">$_POST['title']</error>, '">';
echo "<li>{<error descr="Escape untrusted text for its HTML output context.">$_GET['first']</error>}</li><li>{<error descr="Escape untrusted text for its HTML output context.">$_GET['second']</error>}</li>";
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['first']</error> . $unknown . '<p>' . $_GET['second'];
echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['first']</error> . '</p>', '<p>' . $_GET['second'] . refresh();
