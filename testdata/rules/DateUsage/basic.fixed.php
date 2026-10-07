<?php
function stamp($pattern, Clock $clock)
{
    echo date('D, d M');
    echo \date($pattern);
    echo date($pattern );
}
