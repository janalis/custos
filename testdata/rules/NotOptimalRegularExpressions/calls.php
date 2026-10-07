<?php
function callDemo($file, $needle, $name) {
    preg_match(<weak_warning descr="Drop the leading '.*'; it does not change whether the pattern matches.">'/.*\.log/'</weak_warning>, $file);
    preg_match(<weak_warning descr="Drop the trailing '.*'; it does not change whether the pattern matches.">'/tmp-.*/'</weak_warning>, $file);
    preg_match(<weak_warning descr="Drop the leading '.*'; it does not change whether the pattern matches."><weak_warning descr="Drop the trailing '.*'; it does not change whether the pattern matches.">'/.*-.*/'</weak_warning></weak_warning>, $file);
    $n = preg_match_all(<weak_warning descr="Drop the leading '.*'; it does not change whether the pattern matches.">'/.*;/'</weak_warning>, $file);
    preg_match('/.*\.log/', $file, $m);
    preg_match('/^.*\.log/', $file);
    preg_match('/.*=\0/', $file);
    preg_replace('/tmp-.*/', '', $file);

    <warning descr="Pass the delimiter to preg_quote() so it is escaped as well.">preg_quote('#a.b#')</warning>;
    preg_quote('#a.b#', '#');
    preg_quote($needle);
    preg_quote('a.b');

    if (<weak_warning descr="Use preg_match() when only testing for a match.">preg_match_all</weak_warning>('/\d+/', $file)) {}
    $n = preg_match_all('/\d+/', $file);
    if (preg_match_all('/\d+/', $file, $all)) {}

    preg_match('/^[a-z]+$/', <warning descr="Drop the case conversion and add the /i flag instead.">strtolower($name)</warning>);
    preg_match('/^[a-z]+$/i', <warning descr="The case conversion is redundant: the pattern is already case-insensitive.">mb_strtoupper($name)</warning>);
    preg_match('/^[a-z]+$/', strtolower($name), $m);
    preg_match('/^[a-z]+$/', trim($name));
}
