<?php
// presets: laravel per; risky
class A
{
    use \Alpha, beta, Zed;
    use C\One,
        C\Two {
        C\One::f insteadof C\Two;
    }
    use B /* keep */, A;
    use Sorted, Uns;
}
