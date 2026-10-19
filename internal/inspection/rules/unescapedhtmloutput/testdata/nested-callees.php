<?php
// A function of the file is looked into a few calls deep: what it calls must spare the request tables too.
function refill() { $_GET['id'] = $_POST['id']; }
function audit($m) { error_log($m); }
function record($m) { audit($m); }
function store($m) { refill(); }
function cleanDeep() { $_GET['id'] = (int) $_GET['id']; record('x'); echo '<p>' . $_GET['id'] . '</p>'; }
function refilledDeep() { $_GET['id'] = (int) $_GET['id']; store('x'); echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['id']</error> . '</p>'; }
