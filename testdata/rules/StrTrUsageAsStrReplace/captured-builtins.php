<?php
namespace Text {
    function str_replace($search, $replace, $subject) { return $subject; }

    function slug($title) {
        return <weak_warning descr="Use '\str_replace('-', '_', $title)' instead.">strtr($title, '-', '_')</weak_warning>;
    }
}
