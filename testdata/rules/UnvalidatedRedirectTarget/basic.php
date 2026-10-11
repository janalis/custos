<?php
<error descr="Validate the redirect target against the permitted destination policy.">header('Location: ' . $_GET['next'])</error>;
// Request data stays a redirect target whatever follows it.
<error descr="Validate the redirect target against the permitted destination policy.">header('Location: ' . $_GET['return'] . '?saved=1')</error>;
<error descr="Validate the redirect target against the permitted destination policy.">header("Location: {$_GET['return']}")</error>;
<error descr="Validate the redirect target against the permitted destination policy.">header(('Location: ' . $_GET['return']), true, 302)</error>;
// A local path prefix does not confine an unencoded request value.
<error descr="Validate the redirect target against the permitted destination policy.">header('Location: /search?q=' . $_GET['q'])</error>;
<error descr="Validate the redirect target against the permitted destination policy.">header('Location: /search?q=' . urlencode($_GET['q']) . '&back=' . trim($_GET['back']))</error>;
<error descr="Validate the redirect target against the permitted destination policy.">header('Location: //search?q=' . urlencode($_GET['q']))</error>;
// A selection carries the taint of the results it can produce.
<error descr="Validate the redirect target against the permitted destination policy.">header('Location: ' . match ($_GET['kind']) { 'back' => $_GET['return'], default => '/home' })</error>;
$chosen = match ($_GET['kind']) { 'back' => $_GET['return'], default => '/home' };
<error descr="Validate the redirect target against the permitted destination policy.">header('Location: ' . $chosen)</error>;
