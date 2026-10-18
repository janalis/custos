<?php
// Native transforms whose output contract proves or keeps escaping.
function digits() { $id = $_GET['id']; if (!ctype_digit($id)) { return; } echo '<p>' . $id . '</p>'; }
function formatted() { echo '<p>' . number_format($_GET['n'], 2) . '</p>'; }
function separators() { echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">number_format($_GET['n'], 2, $_GET['d'], '.')</error> . '</p>'; }
function validated() { echo '<p>' . filter_var($_GET['n'], FILTER_VALIDATE_INT) . '</p>'; }
function sanitized() { echo '<a title="' . filter_var($_GET['t'], FILTER_SANITIZE_SPECIAL_CHARS) . '">x</a>'; }
function raw() { echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">filter_var($_GET['t'])</error> . '</p>'; }
function spreadFlags(array $o) { echo '<a title="' . <error descr="Escape untrusted text for its HTML output context.">htmlspecialchars($_GET['t'], ...$o)</error> . '">x</a>'; }
function breaks() { echo '<a title="' . nl2br(htmlspecialchars($_GET['t'], ENT_QUOTES)) . '">x</a>'; }
