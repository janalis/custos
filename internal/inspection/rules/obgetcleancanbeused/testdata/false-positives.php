<?php
function renderPanel($tpl, $keep, $cached, $buf) {
    ob_start();
    echo ob_get_contents();
    ob_end_clean();

    ob_start();
    $copy = ob_get_contents();
    flush_logs();
    ob_end_clean();

    ob_start();
    $last = ob_get_contents();
    $ok = ob_end_clean();

    ob_start();
    $doc = ob_get_contents();
    /** separator */
    ob_end_clean();

    ob_start();
    $q = Other\ob_get_contents();
    ob_end_clean();

    ob_start();
    $m = $buf->ob_get_contents();
    ob_end_clean();

    ob_start();
    $s = ob_get_contents();
    @ob_end_clean();

    ob_start();
    $t = $keep ? ob_get_contents() : '';
    ob_end_clean();

    ob_start();
    $u = $keep ?: ob_get_contents();
    ob_end_clean();

    ob_start();
    $keep && $v = ob_get_contents();
    ob_end_clean();

    ob_start();
    $keep || $v = ob_get_contents();
    ob_end_clean();

    ob_start();
    $keep and $v = ob_get_contents();
    ob_end_clean();

    ob_start();
    $keep or $v = ob_get_contents();
    ob_end_clean();

    ob_start();
    $w = $cached ?? ob_get_contents();
    ob_end_clean();

    ob_start();
    $cached ??= ob_get_contents();
    ob_end_clean();

    ob_start();
    $x = match ($keep) { true => ob_get_contents(), default => '' };
    ob_end_clean();

    ob_start();
    $y = match (true) { $keep, ob_get_contents() === '' => 1, default => 0 };
    ob_end_clean();

    ob_start();
    $buf?->store(ob_get_contents());
    ob_end_clean();

    ob_start();
    $buf?->sink->store(ob_get_contents());
    ob_end_clean();

    ob_start();
    $z = $buf?->parts[ob_get_contents()];
    ob_end_clean();

    ob_start();
    $z = $buf?->{ob_get_contents()};
    ob_end_clean();

    ob_start();
    $reader = ob_get_contents(...);
    ob_end_clean();

    ob_start();
    $now = ob_get_contents();
    ob_end_clean(...);

    return [$copy, $last, $ok, $doc, $q, $m, $s, $t, $u, $v, $w, $x, $y, $z, $reader, $now];
}
