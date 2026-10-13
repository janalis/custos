<?php

namespace Billing\Models {
    class Invoice {}
}

namespace Model {
    class ModelRegistry {}
}

namespace Billing\Http {
    use Billing\Models;
    use Model as M;
    use billing\Models as Wrong;

    class Gateway {}

    return [
        // The namespace keyword stands for the current namespace.
        namespace\Gateway::class,
        NAMESPACE\Gateway::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">namespace\gateway</error>::class,
        // Only the first segment goes through the import, in any case.
        Models\Invoice::class,
        models\Invoice::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">Models\invoice</error>::class,
        // The alias text recurring in the class name is left alone.
        M\ModelRegistry::class,
        m\ModelRegistry::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">M\modelregistry</error>::class,
        // A wrongly cased import target ends up in the string.
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">Wrong\Invoice</error>::class,
    ];
}

namespace {
    use Billing\Models as Bm;

    return [
        Bm\Invoice::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">Bm\INVOICE</error>::class,
        billing\models\invoice::class,
        namespace\arrayobject::class,
    ];
}
