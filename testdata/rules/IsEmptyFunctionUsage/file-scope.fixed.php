<?php
$when = new DateTime();
if ($when === null) {
    echo 'never';
}
/**
 * @var DateTime $conf set by the including page
 */
if (empty($conf) || !is_object($conf)) {
    exit(1);
}
if ($_GET['o'] == 'x') {
    $src = new DateTime();
}
if (!empty($src)) {
    echo 2;
}
$check = fn (?DateTime $d) => $d === null;
