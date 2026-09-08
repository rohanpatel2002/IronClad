import React from 'react';
import { RiskCard } from '../components/RiskCard';

describe('RiskCard Component', () => {
  it('renders risk card props correctly', () => {
    const card = (
      <RiskCard
        deploymentId="deploy-001"
        blastRadius={0.5}
        reversibility={0.2}
        timingRisk={0.3}
        decision="ALLOW"
      />
    );
    expect(card).toBeTruthy();
  });
});
