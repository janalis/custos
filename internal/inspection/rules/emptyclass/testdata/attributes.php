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
    class Tagged {}
}
