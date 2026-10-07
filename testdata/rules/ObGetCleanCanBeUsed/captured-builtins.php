<?php
namespace View {
    function ob_get_clean() { return ''; }

    function render($tpl) {
        ob_start();
        include $tpl;
        $html = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
        ob_end_clean();
        return $html;
    }
}
