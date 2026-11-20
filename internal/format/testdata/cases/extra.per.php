<?php
// presets: per
class A
{
    #[Deprecated]
    const X = 1;
    // note
}
#[Attr]
const Y = 2;
Foo::bar(1);
for ($i = f(1), $j = [2]; $i < g(3); $i++, $j[] = 1) {}
$s = "a {$b["c"]} {$d->e("f")} {$g[<<<EOT
x
EOT]}";
foo($a, $b);
$x = [1, 2];
match ($a) { 1 => 2 };
if ($a) {
} elseif ($b) {
} else /* c */ if ($c) {
}
