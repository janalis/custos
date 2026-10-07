<?php
class Repo { public static ?object $current = null; }
function describe() {
    if (repo::$current === null) {
        return '';
    }
    return get_class(Repo::$current);
}
