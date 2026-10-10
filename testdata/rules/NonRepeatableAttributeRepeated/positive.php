<?php
#[Attribute] class Label {} #[Label, <error descr="Remove the repeated attribute.">Label</error>] class Parcel {}
