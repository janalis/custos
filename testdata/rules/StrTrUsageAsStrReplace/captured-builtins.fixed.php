<?php
namespace Text {
    function str_replace($search, $replace, $subject) { return $subject; }

    function slug($title) {
        return \str_replace('-', '_', $title);
    }
}
