<?php
function named($file) {
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir(directory: dirname($file), recursive: true) && !is_dir(...)) { ... }'.">mkdir(directory: dirname($file), recursive: true);</error>
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir(recursive: true, directory: $file . '/a') && !is_dir(...)) { ... }'.">mkdir(recursive: true, directory: $file . '/a');</error>
}

function strings($base, $seq, $name) {
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir(&quot;{$base}/job-{$seq->next()}&quot;) && !is_dir(...)) { ... }'.">mkdir("{$base}/job-{$seq->next()}");</error>
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir(&quot;$seq->dir/a&quot;) && !is_dir(...)) { ... }'.">mkdir("$seq->dir/a");</error>
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir(&quot;{$base['k']}/a&quot;) && !is_dir(...)) { ... }'.">mkdir("{$base['k']}/a");</error>
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir(&quot;${$name}/a&quot;) && !is_dir(...)) { ... }'.">mkdir("${$name}/a");</error>
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir($$name) && !is_dir(...)) { ... }'.">mkdir($$name);</error>
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir(`pwd`) && !is_dir(...)) { ... }'.">mkdir(`pwd`);</error>
    <error descr="mkdir() outcome is ignored; use 'if (!mkdir(42) && !is_dir(...)) { ... }'.">mkdir(42);</error>
}
