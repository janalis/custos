<?php
function mirroredExit(string $dir, bool $quiet) { $root = realpath($dir); if (!$root && $quiet) { return null; } return <warning descr="Reject canonicalization failure before joining paths.">is_file($root . '/a.json')</warning>; }
function mirroredBranch(string $dir, bool $force) { $root = realpath($dir); if ($root || $force) { return <warning descr="Reject canonicalization failure before joining paths.">is_file($root . '/a.json')</warning>; } return null; }
function asserted(string $dir) { $root = realpath($dir); assert($root !== false); return <warning descr="Reject canonicalization failure before joining paths.">is_file($root . '/a.json')</warning>; }
function staleCheck(string $dir, string $other) { $root = realpath($dir); if (!$root) { return null; } $root = realpath($other); return <warning descr="Reject canonicalization failure before joining paths.">is_file($root . '/a.json')</warning>; }
function wrongArm(string $dir) { $root = realpath($dir); return $root ? null : <warning descr="Reject canonicalization failure before joining paths.">is_file($root . '/a.json')</warning>; }
