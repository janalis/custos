<?php // presets: per; set: import_order=length, remove_unused_imports=true
use Long\Name\Here, Short\N;
use Alias\Thing as T;
use Unused\Alias as U;
use Doc\Only;
use Qualified\Ns;
use \Lead\Slash;
// stop
use Z\After;

/** @var Only $x */
$t = new T(Ns\Sub::X, Slash::Y, After::Z, N::A, Here::B);
