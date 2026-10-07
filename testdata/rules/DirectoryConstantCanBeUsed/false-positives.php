<?php
function bootstrap()
{
    $up    = dirname(__FILE__, 2);
    $other = dirname(__FILE__ . '/../lib');
    $paren = dirname((__FILE__));
    $none  = dirname();
    $dyn   = $resolver->dirname(__FILE__);
    $stat  = Path::dirname(__FILE__);
    $var   = $fn(__FILE__);
    $dir   = dirname(__DIR__);
    $ns    = Tools\dirname(__FILE__);
}
