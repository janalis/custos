<?php
function renderPanel($tpl) {
    ob_start();
    include $tpl;
    $markup = ob_get_clean();
    

    ob_start();
    show_footer();
    $size = strlen(\ob_get_clean());
    // drop the buffer
    

    ob_start();
    process(ob_get_contents(), fn() => ob_get_contents());
    \ob_end_clean();

    return [$markup, $size];
}
