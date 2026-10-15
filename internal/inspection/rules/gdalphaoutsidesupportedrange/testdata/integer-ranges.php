<?php
function invalidChoices($image, bool $pick): void {
    $alpha = $pick ? 128 : 200;
    <error descr="Supply GD alpha between zero and 127.">imagecolorallocatealpha($image, 1, 2, 3, $alpha)</error>;
}
function crossesBoundary($image, bool $pick): void {
    $alpha = $pick ? 127 : 128;
    imagecolorallocatealpha($image, 1, 2, 3, $alpha);
}
function negativeChoices($image, bool $pick): void {
    $alpha = $pick ? -4 : -2;
    <error descr="Supply GD alpha between zero and 127.">imagecolorclosestalpha($image, 1, 2, 3, $alpha)</error>;
}
function boundedLoop($image): void {
    for ($alpha = 130; $alpha > 127; $alpha--) {
        <error descr="Supply GD alpha between zero and 127.">imagecolorallocatealpha($image, 1, 2, 3, $alpha)</error>;
    }
}
function overflowingAlpha($image): void {
    $alpha = 9223372036854775807;
    $alpha++;
    imagecolorallocatealpha($image, 1, 2, 3, $alpha);
}
