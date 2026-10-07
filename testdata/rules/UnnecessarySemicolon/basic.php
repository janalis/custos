<?php
namespace App {
    function tidy() {
        prepare();<weak_warning descr="Stray semicolon; remove it.">;</weak_warning>
        <weak_warning descr="Stray semicolon; remove it.">;</weak_warning>
        return;
    }
    class Box {}<weak_warning descr="Stray semicolon; remove it.">;</weak_warning>
    interface Shape {
        public function area();<weak_warning descr="Stray semicolon; remove it.">;</weak_warning>
    }
    switch ($mode) {
        case 1:
            run();<weak_warning descr="Stray semicolon; remove it.">;</weak_warning>
            break;
    }
    if ($alt):
        <weak_warning descr="Stray semicolon; remove it.">;</weak_warning>
    endif;
}
?>
<p><?= $title<weak_warning descr="Stray semicolon; remove it.">;</weak_warning> ?></p>
<p><?= $title <weak_warning descr="Stray semicolon; remove it.">;</weak_warning>?></p>
<p><?= $title<weak_warning descr="Stray semicolon; remove it.">;</weak_warning> /* note */ ?></p>
