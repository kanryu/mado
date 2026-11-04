struct VSOutput {
    float4 position : SV_Position;
    float4 color : COLOR0;
};

VSOutput main(uint vertexID : SV_VertexID) {
    float2 positions[3] = {
        float2(0.0, -0.65),
        float2(-0.65, 0.65),
        float2(0.65, 0.65)
    };
    float4 colors[3] = {
        float4(1.0, 0.15, 0.05, 1.0),
        float4(1.0, 0.55, 0.05, 1.0),
        float4(0.95, 0.05, 0.20, 1.0)
    };
    if (vertexID >= 3) {
        float2 linePositions[6] = {
            float2(0.82, -0.82), float2(0.55, -0.45),
            float2(0.55, -0.45), float2(0.98, -0.45),
            float2(0.98, -0.45), float2(0.82, -0.82)
        };
        VSOutput lineOutput;
        lineOutput.position = float4(linePositions[vertexID - 3], 0.0, 1.0);
        lineOutput.color = float4(0.2, 0.9, 1.0, 1.0);
        return lineOutput;
    }
    VSOutput output;
    output.position = float4(positions[vertexID], 0.0, 1.0);
    output.color = colors[vertexID];
    return output;
}
