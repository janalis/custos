<?php
// presets: per; set: import_order=length, remove_unused_imports=true
use Short\N;
use Doc\Only;
use Lead\Slash;
use Alias\Thing as T;
use Qualified\Ns;
use Long\Name\Here;
// stop
use Z\After;

/** @var Only $x */
$t = new T(Ns\Sub::X, Slash::Y, After::Z, N::A, Here::B);
