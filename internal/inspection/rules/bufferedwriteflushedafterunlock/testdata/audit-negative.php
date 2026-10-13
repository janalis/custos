<?php
$h=tmpfile(); flock($h,LOCK_EX); fwrite($h,'data'); $alias=$h; fflush($alias); flock($h,LOCK_UN); fflush($h); fclose($h);

function earlier_alias($h) {
    $alias = $h;
    flock($h, LOCK_EX);
    fwrite($h, 'data');
    fflush($alias);
    flock($h, LOCK_UN);
    fflush($h);
}
