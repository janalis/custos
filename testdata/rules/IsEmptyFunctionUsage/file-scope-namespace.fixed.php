<?php
namespace App {
    $when = new \DateTime();
    if ($when === null) {
        echo 'never';
    }
    /** @var \DateTime $page */
    if (empty($page)) {
        echo 'unset';
    }
}
