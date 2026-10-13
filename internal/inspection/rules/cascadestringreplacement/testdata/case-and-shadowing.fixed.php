<?php
namespace App {
    function str_replace($s, $r, $subject) { return $subject; }

    function slug($title) {
        $title = str_replace('&', 'and', $title);
        return str_replace(' ', '-', $title);
    }
}

namespace {
    function slug($title) {
        return str_replace(array('&', ' '), array('and', '-'), $title);
    }
}
