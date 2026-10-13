<?php
namespace App;

function ob_get_contents() { return ''; }

function render() {
    ob_start();
    $html = ob_get_contents();
    ob_end_clean();

    ob_start();
    $real = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">\ob_get_contents()</warning>;
    ob_end_clean();
    return [$html, $real];
}
