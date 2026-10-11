<?php
function review_dropped_comment($h) {
    if (<warning descr="Compare the directory entry with false.">! /* end of dir? */ readdir($h)</warning>) { return 1; }
    if (<warning descr="Compare the directory entry with false.">readdir($h) /* loose on purpose */ == false</warning>) { return 2; }
    if (<warning descr="Compare the directory entry with false.">false == # leading
        readdir($h)</warning>) { return 3; }
    if (<warning descr="Compare the directory entry with false.">! // negated
        (readdir($h))</warning>) { return 4; }
    return 0;
}
