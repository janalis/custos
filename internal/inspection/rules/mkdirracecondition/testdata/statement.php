<?php
function prepare($root, $cfg) {
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir($root, 0750) && !is_dir(...)) { ... }'.">@mkdir($root, 0750);</error>
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir($cfg->dir()) && !is_dir(...)) { ... }'.">mkdir($cfg->dir());</error>
}
