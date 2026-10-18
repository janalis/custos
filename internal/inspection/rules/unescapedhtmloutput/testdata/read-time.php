<?php
// A replacement clears only what is read from the table after it.
function copiedBefore() { $v = $_GET['v']; $_GET['v'] = 'safe'; echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$v</error>; }
function copiedBeforeEmptied() { $v = $_GET['v']; $_GET = []; echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$v</error>; }
function readAfter() { $_GET['v'] = 'safe'; $v = $_GET['v']; echo '<p>' . $v; }
function escapedBefore() { $v = htmlspecialchars($_GET['v']); $_GET = []; echo '<p>' . $v; }
