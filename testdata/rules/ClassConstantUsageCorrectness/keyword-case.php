<?php
namespace Billing {
    class Invoice {}
}

namespace App {
    use Billing\invoice;
    use Billing\Invoice as Bill;

    class Report
    {
        public function names()
        {
            return [
                <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">invoice</error>::CLASS,
                <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">BILL</error>::Class,
                Bill::CLASS,
                SELF::CLASS,
                Static::class,
            ];
        }
    }
}
