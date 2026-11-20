<?php // presets: laravel per; risky
class A
{
    use Zed, \Alpha, beta;
    use C\Two,
        C\One {
        C\One::f insteadof C\Two;
    }
    use B /* keep */, A;
    use Sorted, Uns;
}
