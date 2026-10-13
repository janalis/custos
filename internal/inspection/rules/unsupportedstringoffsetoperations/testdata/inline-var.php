<?php
function merge(string $ids, $repo) {
    /** @var string $ids annotates the explode() statement only */
    $parts = explode(',', $ids);
    $ids = [];
    foreach ($parts as $id) {
        $ids[] = (int) $id;
    }
    return $ids;
}
