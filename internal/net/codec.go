package net

import "github.com/Advik-B/Golem/nbt"

// A minimal, valid dimension codec required by the client to log in.
const dimensionCodecSNBT = `
{
   "minecraft:dimension_type": {
      "type": "minecraft:dimension_type",
      "value": [
         {
            "name": "minecraft:overworld",
            "id": 0,
            "element": {
               "ultrawarm": false,
               "natural": true,
               "piglin_safe": false,
               "respawn_anchor_works": false,
               "bed_works": true,
               "has_raids": true,
               "has_skylight": true,
               "has_ceiling": false,
               "fixed_time": 0,
               "min_y": -64,
               "height": 384,
               "logical_height": 384,
               "monster_spawn_light_level": 0,
               "monster_spawn_block_light_limit": 0,
               "infiniburn": "#minecraft:infiniburn_overworld",
               "effects": "minecraft:overworld"
            }
         }
      ]
   },
   "minecraft:worldgen/biome": {
      "type": "minecraft:worldgen/biome",
      "value": [
         {
            "name": "minecraft:plains",
            "id": 0,
            "element": {
               "precipitation": "rain",
               "downfall": 0.4,
               "effects": {
                  "sky_color": 7907327,
                  "fog_color": 12638463,
                  "water_color": 4159204,
                  "water_fog_color": 329011
               },
               "temperature": 0.8
            }
         }
      ]
   }
}`

var dimensionCodec *nbt.CompoundTag

func init() {
	var err error
	dimensionCodec, err = nbt.ParseSNBT(dimensionCodecSNBT)
	if err != nil {
		panic("failed to parse hardcoded dimension codec: " + err.Error())
	}
}

// GetDimensionCodec returns the parsed dimension codec NBT tag.
func GetDimensionCodec() *nbt.CompoundTag {
	return dimensionCodec
}
