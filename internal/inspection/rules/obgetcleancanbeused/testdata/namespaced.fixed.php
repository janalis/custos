<?php
namespace App;

function ob_get_contents() { return ''; }

function render() {
    ob_start();
    $html = ob_get_contents();
    ob_end_clean();

    ob_start();
    $real = \ob_get_clean();
    return [$html, $real];
}
