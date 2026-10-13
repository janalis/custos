<?php
namespace App {
    $when = new \DateTime();
    if (<weak_warning descr="Replace with '$when === null'.">empty($when)</weak_warning>) {
        echo 'never';
    }
    /** @var \DateTime $page */
    if (empty($page)) {
        echo 'unset';
    }
}
