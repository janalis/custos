<?php
$when = new DateTime();
if (<weak_warning descr="Replace with '$when === null'.">empty($when)</weak_warning>) {
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
$check = fn (?DateTime $d) => <weak_warning descr="Replace with '$d === null'.">empty($d)</weak_warning>;
