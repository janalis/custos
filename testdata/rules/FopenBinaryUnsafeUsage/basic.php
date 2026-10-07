<?php
function open_files($path)
{
    $access = 'a';
    $h1 = fopen($path, <warning descr="Add the 'b' flag to the mode for binary-safe file access.">$access</warning>);
    $h2 = fopen($path, <warning descr="Add the 'b' flag to the mode for binary-safe file access.">"r+"</warning>);
    $h3 = fopen($path, <warning descr="Use the 'b' flag instead of 't' for binary-safe file access.">'at'</warning>);
    $h4 = \fopen($path, <error descr="Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+').">'bx'</error>);
    $h5 = fopen($path, <warning descr="Use the 'b' flag instead of 't' for binary-safe file access.">'wt+'</warning>);
    $h6 = fopen($path, <error descr="Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+').">'rbt'</error>);
}
