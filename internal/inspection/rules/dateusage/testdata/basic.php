<?php
function stamp($pattern, Clock $clock)
{
    echo date('D, d M', <weak_warning descr="Redundant time() argument: date() uses the current time by default.">time()</weak_warning>);
    echo \date($pattern, <weak_warning descr="Redundant time() argument: date() uses the current time by default.">\time()</weak_warning>);
    echo date($pattern ,  <weak_warning descr="Redundant time() argument: date() uses the current time by default.">time()</weak_warning> );
}
