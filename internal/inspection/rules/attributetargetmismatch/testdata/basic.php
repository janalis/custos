<?php
#[Attribute(Attribute::TARGET_METHOD)] class Marker {} <error descr="Use the attribute on a permitted target.">#[Marker]</error> class Product {}
