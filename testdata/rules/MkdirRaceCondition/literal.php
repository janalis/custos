<?php
function make() {
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir('/tmp/x') && !is_dir(...)) { ... }'.">mkdir('/tmp/x');</error>
}
