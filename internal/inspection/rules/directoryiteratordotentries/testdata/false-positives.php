<?php
function onlyFiles($dir) { foreach (new DirectoryIterator($dir) as $e) { if ($e->isFile()) { unlink($e->getPathname()); } } }
function notFileContinue($dir) { foreach (new DirectoryIterator($dir) as $e) { if (!$e->isFile()) { continue; } unlink($e->getPathname()); } }
function dotOrDir($dir) { foreach (new DirectoryIterator($dir) as $e) { if ($e->isDot() || $e->isDir()) { continue; } unlink($e->getPathname()); } }
function notDotAnd($dir) { foreach (new DirectoryIterator($dir) as $e) { if (!$e->isDot() && $e->isFile()) { unlink($e->getPathname()); } } }
function dotNames($dir) { foreach (new DirectoryIterator($dir) as $e) { if ($e->getFilename() === '.' || $e->getFilename() === '..') { continue; } unlink($e->getPathname()); } }
function listedNames($dir) { foreach (new DirectoryIterator($dir) as $e) { if (in_array($e->getFilename(), array('.', '..'), true)) { continue; } unlink($e->getPathname()); } }
function elseBranch($dir) { foreach (new DirectoryIterator($dir) as $e) { if ($e->isDot()) { echo ''; } else { unlink($e->getPathname()); } } }
function upperCase($dir) { foreach (new DirectoryIterator($dir) as $e) { if ($e->ISDOT()) { continue; } unlink($e->getPathname()); } }
function parens($dir) { foreach (new DirectoryIterator($dir) as $e) { if (($e->isDot())) { continue; } unlink($e->getPathname()); } }
function shortCircuit($dir) { foreach (new DirectoryIterator($dir) as $e) { $e->isDot() || unlink($e->getPathname()); } }
