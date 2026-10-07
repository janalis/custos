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
    process(ob_get_clean(), fn() => ob_get_contents());
    

    return [$markup, $size];
}
