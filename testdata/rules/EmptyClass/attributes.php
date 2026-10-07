<?php
namespace App\Queue {
    use Attribute;

    #[Attribute(Attribute::TARGET_CLASS)]
    final class Unique {}

    #[\Attribute]
    class Retry {}

    #[\JetBrains\Immutable, \ATTRIBUTE]
    class Pinned {}

    #[Marker]
    class <weak_warning descr="This class declares no members; remove it or give it a purpose.">Tagged</weak_warning> {}
}
