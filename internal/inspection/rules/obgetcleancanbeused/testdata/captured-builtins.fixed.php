<?php
namespace View {
    function ob_get_clean() { return ''; }

    function render($tpl) {
        ob_start();
        include $tpl;
        $html = \ob_get_clean();
        
        return $html;
    }
}
