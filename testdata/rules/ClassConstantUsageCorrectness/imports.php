<?php

namespace Shop\Deep\Parts {
    class Bolt {}
    class Nut {}
}

namespace Garage {
    use Shop\Deep\Parts\Nut;
    use Shop\Deep\Parts\Nut as Fastener;
    use Shop\deep\Parts as P;
    use Shop\Deep\Parts as Q;

    return [
        Nut::class,
        Fastener::class,
        nut::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">P\Bolt</error>::class,
        Q\Bolt::class,
    ];
}
