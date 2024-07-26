// param $resources is a list of newResourceParam structs with the fields 
UNWIND $resources AS resource
MERGE (r:Resource {type: resource.label, id: resource.id})
  ON CREATE
    SET
      r.created = timestamp(),
      r += resource.props
  ON MATCH
    SET
      r.lastSeen = timestamp(),
      r += resource.props
FOREACH (contained_node IN resource.contained_in |
  MERGE (r)-[:IS_CONTAINED]->(c:Resource {type: contained_node.label, id: contained_node.id})
)
