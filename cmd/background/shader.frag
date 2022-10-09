#version 410 core
in vec2 TexCoord;
out vec4 color;
uniform float ScreenWidth;
uniform float ScreenHeight;

float abs(float a) {
	return a > 0 ? a : -a;
}
float pdist(float a, float b){
	return abs((a) - (b));
}
float dist(vec2 a, vec2 b){
	vec2 d = vec2(pdist(a.x, b.x), pdist(a.y, b.y));
	return sqrt(d.x * d.x + d.y * d.y);
}
void main()
{
	float ratio = ScreenHeight / ScreenWidth;
	vec2 norm = vec2(TexCoord.x, TexCoord.y * ratio);
	vec2 r = vec2(0.0, 0.0);
	float d = dist(norm, r);
	if (d < 0.022) {
		color = vec4(0.3f, 0.0f, 0.2f, 1.0f);
		return;
	}
    color = vec4(0.5f, 0.1f, 0.55f, 1.0f) * (0.2 / d);
}

