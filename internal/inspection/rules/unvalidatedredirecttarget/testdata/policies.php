<?php
// A match over request data with literal results is a finite destination policy.
$page = match ($_GET['next']) { 'cart' => '/cart', default => '/home' };
header('Location: ' . $page);
header('Location: ' . match ($_GET['next']) { 'cart' => '/cart', default => '/home' } . '?saved=1');
header("Location: {$page}");
// A fixed local path whose query values are URL-encoded stays on this origin.
header('Location: /search?q=' . urlencode($_GET['q']));
header('Location: /search?q=' . rawurlencode($_GET['q']) . '&page=2');
header('Location: /search?' . http_build_query(['q' => $_GET['q']]));
$query = urlencode($_GET['q']);
header("Location: /search?q={$query}");
// Other headers and non-text first arguments are not redirect targets here.
header("X-Trace: {$_GET['trace']}");
header('X-Trace: ' . $_GET['trace'] . '; done');
header(`printf 'Location: ' {$_GET['next']}`);
header(...$_GET['headers']);
header($_GET['raw']);
