<?php
// Writes that run before the output reads a request table, in the order they run.
function refill() { $_GET['id'] = $_POST['id']; }
function replacedEachTurn($c) { while ($c) { $_GET['id'] = (int) $_GET['id']; echo '<p>' . $_GET['id'] . '</p>'; refill(); } }
function refilledEachTurn($c) { $_GET['id'] = (int) $_GET['id']; while ($c) { echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['id']</error> . '</p>'; refill(); } }
function refilledInsideOutput() { $_GET['id'] = (int) $_GET['id']; echo '<p>' . refill() . <error descr="Escape untrusted text for its HTML output context.">$_GET['id']</error> . '</p>'; }
function refilledThroughCallback() { $_GET['id'] = (int) $_GET['id']; array_map('refill', [1]); echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$_GET['id']</error> . '</p>'; }
function readBeforeLoop($c) { $id = $_GET['id']; $_GET['id'] = 1; while ($c) { echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$id</error> . '</p>'; refill(); } }
