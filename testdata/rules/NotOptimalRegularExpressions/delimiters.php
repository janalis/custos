<?php
function delimiters($line) {
    preg_match(<warning descr="Pattern has no valid delimiters.">'xa.bxe'</warning>, $line);
    preg_match(<warning descr="Pattern has no valid delimiters.">'1k\d1'</warning>, $line);
    preg_match(<warning descr="Pattern has no valid delimiters.">'\k\\'</warning>, $line);
    preg_match(<warning descr="Pattern has no valid delimiters.">'  '</warning>, $line);
    preg_match(' /k\d/', $line);
    preg_match('  #k\d#i', $line);
    preg_quote('xa.bx');
    <warning descr="Pass the delimiter to preg_quote() so it is escaped as well.">preg_quote(' #a.b#')</warning>;
}
