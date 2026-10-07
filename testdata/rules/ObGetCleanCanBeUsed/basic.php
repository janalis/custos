<?php
function renderPanel($tpl) {
    ob_start();
    include $tpl;
    $markup = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    ob_start();
    show_footer();
    $size = strlen(<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">\ob_get_contents()</warning>);
    // drop the buffer
    ob_end_clean();

    ob_start();
    process(<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>, fn() => ob_get_contents());
    \ob_end_clean();

    return [$markup, $size];
}
