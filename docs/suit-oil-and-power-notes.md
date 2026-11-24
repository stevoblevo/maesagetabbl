# Suit notes: massage oil, cooling, warming, power

## Massage oil: yes, with a spec
- Oil is not a conductivity problem. Mineral oil is a dielectric (relative permittivity about 2–2.4); sweat and water (about 80) are the real shorting risk.
- Correction to an earlier claim: spandex/elastane resists body oils and lotions. Oils do not break it down.
- What oil does affect: grip on capacitive and touch sensing, how well the fabric wicks, and the life of skin-facing adhesives and seams.
- Spec direction: low-viscosity, non-conductive, electronics-safe oil (mineral oil, or a dielectric fluid of our own spec); an oil-shedding outer face over sealed potting around the electronics; electronics behind a gasket rated for both oil and sweat.

## Piezo: sensing and haptics, not power
- Harvesting from walking: a Penn State footwear harvester averaged about 7–9 mW per shoe at 3 mph (2018). The suit's motors draw hundreds of milliwatts to watts, so harvesting can't charge it.
- Cooling: piezoMEMS "fan on a chip" (xMEMS µCooling) moves air in a 1 mm package. This is useful for venting electronics pockets.
- Cooling and warming on skin: a Peltier module (as in the Sony Reon Pocket Pro) does both.
- Plain warming: a resistive heater is the efficient option.
- Best piezo uses: pressure and touch sensing, and crisp haptic actuation.
